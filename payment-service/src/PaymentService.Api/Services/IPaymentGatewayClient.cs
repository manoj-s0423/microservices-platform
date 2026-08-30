namespace PaymentService.Api.Services;

public record GatewayChargeRequest(string OrderId, int AmountCents, string Currency);

public record GatewayChargeResult(bool Approved, string? TransactionId, string? DeclineReason);

/// <summary>
/// Abstraction over the external card-processing gateway (e.g. Stripe,
/// Braintree, a bank's own processor). This is the platform's one
/// external, third-party API dependency - everything else in ShopStream
/// is an internal microservice.
/// </summary>
public interface IPaymentGatewayClient
{
    Task<GatewayChargeResult> ChargeAsync(GatewayChargeRequest request, CancellationToken cancellationToken);
}
