namespace PaymentService.Api.Models;

public enum PaymentStatus
{
    Pending,
    Succeeded,
    Failed,
    Declined
}

/// <summary>
/// A single charge attempt against an order. IdempotencyKey is unique and
/// set to the order-service's order id, so a retried request (from
/// order-service's own retry policy or a client-side retry after a
/// timeout) never results in a double charge - the second request finds
/// the existing row and returns its result instead of charging again.
/// </summary>
public class Payment
{
    public Guid Id { get; set; }
    public Guid OrderId { get; set; }
    public Guid UserId { get; set; }
    public int AmountCents { get; set; }
    public string Currency { get; set; } = "USD";
    public PaymentStatus Status { get; set; } = PaymentStatus.Pending;
    public string? TransactionId { get; set; }
    public string? FailureReason { get; set; }
    public string IdempotencyKey { get; set; } = string.Empty;
    public DateTimeOffset CreatedAt { get; set; }
    public DateTimeOffset UpdatedAt { get; set; }
}
