using System.ComponentModel.DataAnnotations;

namespace PaymentService.Api.Models;

public class ChargeRequest
{
    [Required]
    public string OrderId { get; set; } = string.Empty;

    [Required]
    public string UserId { get; set; } = string.Empty;

    [Range(1, int.MaxValue, ErrorMessage = "amountCents must be a positive integer")]
    public int AmountCents { get; set; }

    [Required]
    [StringLength(3, MinimumLength = 3)]
    public string Currency { get; set; } = "USD";
}

public class ChargeResponse
{
    public Guid Id { get; set; }
    public string OrderId { get; set; } = string.Empty;
    public string Status { get; set; } = string.Empty; // SUCCEEDED | FAILED | DECLINED | PENDING
    public string? TransactionId { get; set; }
    public string? Reason { get; set; }
    public int AmountCents { get; set; }
    public string Currency { get; set; } = string.Empty;
    public DateTimeOffset CreatedAt { get; set; }

    public static ChargeResponse FromPayment(Payment payment) => new()
    {
        Id = payment.Id,
        OrderId = payment.OrderId.ToString(),
        Status = payment.Status.ToString().ToUpperInvariant(),
        TransactionId = payment.TransactionId,
        Reason = payment.FailureReason,
        AmountCents = payment.AmountCents,
        Currency = payment.Currency,
        CreatedAt = payment.CreatedAt,
    };
}
