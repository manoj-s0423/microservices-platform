using System.Net.Http.Json;
using Microsoft.Extensions.Options;

namespace PaymentService.Api.Services;

/// <summary>
/// "Live" gateway implementation - calls a real external payment
/// processor over HTTP. Registered instead of SimulatedPaymentGatewayClient
/// when Gateway:Mode = "live". Resilience (timeout, retry, circuit
/// breaker) is configured on the HttpClient itself via Polly in
/// Program.cs, not here - this class stays a thin HTTP adapter.
/// </summary>
public class HttpPaymentGatewayClient : IPaymentGatewayClient
{
    private readonly HttpClient _httpClient;
    private readonly ILogger<HttpPaymentGatewayClient> _logger;

    public HttpPaymentGatewayClient(HttpClient httpClient, IOptions<GatewayOptions> options, ILogger<HttpPaymentGatewayClient> logger)
    {
        _httpClient = httpClient;
        _logger = logger;
        if (!string.IsNullOrWhiteSpace(options.Value.BaseUrl))
        {
            _httpClient.BaseAddress = new Uri(options.Value.BaseUrl);
        }
    }

    public async Task<GatewayChargeResult> ChargeAsync(GatewayChargeRequest request, CancellationToken cancellationToken)
    {
        try
        {
            var response = await _httpClient.PostAsJsonAsync("/v1/charges", request, cancellationToken);
            response.EnsureSuccessStatusCode();

            var payload = await response.Content.ReadFromJsonAsync<ExternalGatewayResponse>(cancellationToken: cancellationToken);
            if (payload is null)
            {
                return new GatewayChargeResult(false, null, "empty_gateway_response");
            }

            return new GatewayChargeResult(payload.Approved, payload.TransactionId, payload.DeclineReason);
        }
        catch (TaskCanceledException ex) when (!cancellationToken.IsCancellationRequested)
        {
            _logger.LogError(ex, "payment gateway request timed out");
            throw new PaymentGatewayUnavailableException("Payment gateway timed out", ex);
        }
        catch (HttpRequestException ex)
        {
            _logger.LogError(ex, "payment gateway request failed");
            throw new PaymentGatewayUnavailableException("Payment gateway is unreachable", ex);
        }
    }

    private record ExternalGatewayResponse(bool Approved, string? TransactionId, string? DeclineReason);
}

public class PaymentGatewayUnavailableException : Exception
{
    public PaymentGatewayUnavailableException(string message, Exception inner) : base(message, inner)
    {
    }
}
