using Microsoft.AspNetCore.Mvc;
using PaymentService.Api.Models;
using PaymentService.Api.Services;

namespace PaymentService.Api.Controllers;

[ApiController]
[Route("api/v1/payments")]
public class PaymentsController : ControllerBase
{
    private readonly PaymentProcessingService _paymentService;
    private readonly ILogger<PaymentsController> _logger;

    public PaymentsController(PaymentProcessingService paymentService, ILogger<PaymentsController> logger)
    {
        _paymentService = paymentService;
        _logger = logger;
    }

    [HttpPost]
    public async Task<ActionResult<ChargeResponse>> Charge(
        [FromBody] ChargeRequest request,
        [FromHeader(Name = "Idempotency-Key")] string? idempotencyKeyHeader,
        CancellationToken cancellationToken)
    {
        if (!ModelState.IsValid)
        {
            return ValidationProblem(ModelState);
        }

        // Fall back to orderId itself if the caller didn't send the header -
        // order-service always sends it, but this keeps the endpoint safe
        // to call directly (e.g. from a test client) too.
        var idempotencyKey = string.IsNullOrWhiteSpace(idempotencyKeyHeader) ? request.OrderId : idempotencyKeyHeader;

        try
        {
            var payment = await _paymentService.ChargeAsync(request, idempotencyKey, cancellationToken);
            var response = ChargeResponse.FromPayment(payment);

            return payment.Status switch
            {
                Models.PaymentStatus.Succeeded => Ok(response),
                Models.PaymentStatus.Declined => StatusCode(StatusCodes.Status402PaymentRequired, response),
                _ => StatusCode(StatusCodes.Status502BadGateway, response),
            };
        }
        catch (ArgumentException ex)
        {
            return BadRequest(new { error = "validation_error", message = ex.Message });
        }
    }

    [HttpGet("{id:guid}")]
    public async Task<ActionResult<ChargeResponse>> GetById(Guid id, CancellationToken cancellationToken)
    {
        var payment = await _paymentService.GetByIdAsync(id, cancellationToken);
        if (payment is null)
        {
            return NotFound(new { error = "payment_not_found", message = $"Payment {id} not found" });
        }

        return Ok(ChargeResponse.FromPayment(payment));
    }
}
