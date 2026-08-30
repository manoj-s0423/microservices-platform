using Microsoft.EntityFrameworkCore;
using PaymentService.Api.Data;
using PaymentService.Api.Models;

namespace PaymentService.Api.Services;

public class PaymentProcessingService
{
    private readonly PaymentDbContext _db;
    private readonly IPaymentGatewayClient _gateway;
    private readonly ILogger<PaymentProcessingService> _logger;

    public PaymentProcessingService(PaymentDbContext db, IPaymentGatewayClient gateway, ILogger<PaymentProcessingService> logger)
    {
        _db = db;
        _gateway = gateway;
        _logger = logger;
    }

    /// <summary>
    /// Charges an order. Idempotent on <paramref name="idempotencyKey"/>:
    /// if a payment already exists for that key, its stored result is
    /// returned unchanged and the gateway is never called again - this is
    /// what makes it safe for order-service to retry a charge after an
    /// ambiguous timeout.
    /// </summary>
    public async Task<Payment> ChargeAsync(ChargeRequest request, string idempotencyKey, CancellationToken cancellationToken)
    {
        var existing = await _db.Payments
            .FirstOrDefaultAsync(p => p.IdempotencyKey == idempotencyKey, cancellationToken);
        if (existing is not null)
        {
            _logger.LogInformation("idempotent replay for key {IdempotencyKey}, returning stored result", idempotencyKey);
            return existing;
        }

        if (!Guid.TryParse(request.OrderId, out var orderId) || !Guid.TryParse(request.UserId, out var userId))
        {
            throw new ArgumentException("orderId and userId must be valid GUIDs");
        }

        var payment = new Payment
        {
            Id = Guid.NewGuid(),
            OrderId = orderId,
            UserId = userId,
            AmountCents = request.AmountCents,
            Currency = request.Currency.ToUpperInvariant(),
            Status = PaymentStatus.Pending,
            IdempotencyKey = idempotencyKey,
            CreatedAt = DateTimeOffset.UtcNow,
            UpdatedAt = DateTimeOffset.UtcNow,
        };

        _db.Payments.Add(payment);
        await _db.SaveChangesAsync(cancellationToken);

        try
        {
            var result = await _gateway.ChargeAsync(
                new GatewayChargeRequest(payment.OrderId.ToString(), payment.AmountCents, payment.Currency),
                cancellationToken);

            payment.Status = result.Approved ? PaymentStatus.Succeeded : PaymentStatus.Declined;
            payment.TransactionId = result.TransactionId;
            payment.FailureReason = result.DeclineReason;
        }
        catch (PaymentGatewayUnavailableException ex)
        {
            _logger.LogError(ex, "gateway unavailable while charging order {OrderId}", payment.OrderId);
            payment.Status = PaymentStatus.Failed;
            payment.FailureReason = "gateway_unavailable";
        }

        payment.UpdatedAt = DateTimeOffset.UtcNow;
        await _db.SaveChangesAsync(cancellationToken);

        return payment;
    }

    public async Task<Payment?> GetByIdAsync(Guid id, CancellationToken cancellationToken)
    {
        return await _db.Payments.FirstOrDefaultAsync(p => p.Id == id, cancellationToken);
    }
}
