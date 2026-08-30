namespace PaymentService.Api.Services;

public class GatewayOptions
{
    public const string SectionName = "Gateway";

    /// <summary>"simulated" (default, no external dependency) or "live" (real HTTP calls to BaseUrl).</summary>
    public string Mode { get; set; } = "simulated";

    public string BaseUrl { get; set; } = string.Empty;
    public int TimeoutSeconds { get; set; } = 5;
    public int RetryAttempts { get; set; } = 2;
    public int DeclineThresholdCents { get; set; } = 1_000_000; // $10,000
}
