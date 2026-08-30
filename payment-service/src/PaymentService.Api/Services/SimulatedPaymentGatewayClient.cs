using Microsoft.Extensions.Options;

namespace PaymentService.Api.Services;

/// <summary>
/// Default gateway implementation for local development, CI, and
/// integration tests - no real external dependency required. Used when
/// Gateway:Mode = "simulated" (the default). Deterministic rules make
/// failure scenarios reproducible on demand:
///   - amountCents above Gateway:DeclineThresholdCents -> declined
///   - currency other than USD/EUR/GBP -> declined ("unsupported currency")
/// </summary>
public class SimulatedPaymentGatewayClient : IPaymentGatewayClient
{
    private static readonly string[] SupportedCurrencies = { "USD", "EUR", "GBP" };
    private readonly GatewayOptions _options;
    private readonly ILogger<SimulatedPaymentGatewayClient> _logger;

    public SimulatedPaymentGatewayClient(IOptions<GatewayOptions> options, ILogger<SimulatedPaymentGatewayClient> logger)
    {
        _options = options.Value;
        _logger = logger;
    }

    public async Task<GatewayChargeResult> ChargeAsync(GatewayChargeRequest request, CancellationToken cancellationToken)
    {
        // Simulate realistic gateway latency so timeout/slow-response
        // scenarios can be exercised even without a real external call.
        await Task.Delay(TimeSpan.FromMilliseconds(50), cancellationToken);

        if (!SupportedCurrencies.Contains(request.Currency, StringComparer.OrdinalIgnoreCase))
        {
            _logger.LogWarning("simulated gateway declining unsupported currency {Currency}", request.Currency);
            return new GatewayChargeResult(false, null, "unsupported_currency");
        }

        if (request.AmountCents > _options.DeclineThresholdCents)
        {
            _logger.LogWarning("simulated gateway declining amount {Amount} over threshold {Threshold}",
                request.AmountCents, _options.DeclineThresholdCents);
            return new GatewayChargeResult(false, null, "amount_exceeds_limit");
        }

        var transactionId = $"sim_txn_{Guid.NewGuid():N}";
        return new GatewayChargeResult(true, transactionId, null);
    }
}
