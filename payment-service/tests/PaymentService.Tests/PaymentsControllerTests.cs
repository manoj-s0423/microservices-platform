using FluentAssertions;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;
using PaymentService.Api.Controllers;
using PaymentService.Api.Data;
using PaymentService.Api.Models;
using PaymentService.Api.Services;
using Xunit;

namespace PaymentService.Tests;

/// <summary>
/// Exercises PaymentsController directly (no HTTP pipeline / WebApplicationFactory)
/// against an in-memory EF Core provider and a mocked gateway - this is the
/// "API test" layer: request DTO in, HTTP status + response DTO out.
/// </summary>
public class PaymentsControllerTests
{
    private static (PaymentsController controller, Mock<IPaymentGatewayClient> gateway) BuildController()
    {
        var options = new DbContextOptionsBuilder<PaymentDbContext>()
            .UseInMemoryDatabase(Guid.NewGuid().ToString())
            .Options;
        var db = new PaymentDbContext(options);
        var gateway = new Mock<IPaymentGatewayClient>();
        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);
        var controller = new PaymentsController(service, NullLogger<PaymentsController>.Instance);
        return (controller, gateway);
    }

    [Fact]
    public async Task Charge_Approved_Returns200WithSucceededStatus()
    {
        var (controller, gateway) = BuildController();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(new GatewayChargeResult(true, "txn_1", null));

        var request = new ChargeRequest
        {
            OrderId = Guid.NewGuid().ToString(),
            UserId = Guid.NewGuid().ToString(),
            AmountCents = 1500,
            Currency = "USD",
        };

        var result = await controller.Charge(request, idempotencyKeyHeader: request.OrderId, CancellationToken.None);

        var okResult = result.Result.Should().BeOfType<OkObjectResult>().Subject;
        var body = okResult.Value.Should().BeOfType<ChargeResponse>().Subject;
        body.Status.Should().Be("SUCCEEDED");
    }

    [Fact]
    public async Task Charge_Declined_Returns402()
    {
        var (controller, gateway) = BuildController();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(new GatewayChargeResult(false, null, "card_declined"));

        var request = new ChargeRequest
        {
            OrderId = Guid.NewGuid().ToString(),
            UserId = Guid.NewGuid().ToString(),
            AmountCents = 1500,
            Currency = "USD",
        };

        var result = await controller.Charge(request, idempotencyKeyHeader: request.OrderId, CancellationToken.None);

        var objectResult = result.Result.Should().BeOfType<ObjectResult>().Subject;
        objectResult.StatusCode.Should().Be(StatusCodes.Status402PaymentRequired);
    }

    [Fact]
    public async Task GetById_UnknownId_Returns404()
    {
        var (controller, _) = BuildController();

        var result = await controller.GetById(Guid.NewGuid(), CancellationToken.None);

        result.Result.Should().BeOfType<NotFoundObjectResult>();
    }

    [Fact]
    public async Task Charge_MissingIdempotencyHeader_FallsBackToOrderId()
    {
        var (controller, gateway) = BuildController();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(new GatewayChargeResult(true, "txn_2", null));

        var request = new ChargeRequest
        {
            OrderId = Guid.NewGuid().ToString(),
            UserId = Guid.NewGuid().ToString(),
            AmountCents = 500,
            Currency = "USD",
        };

        var first = await controller.Charge(request, idempotencyKeyHeader: null, CancellationToken.None);
        var second = await controller.Charge(request, idempotencyKeyHeader: null, CancellationToken.None);

        var firstBody = (ChargeResponse)((OkObjectResult)first.Result!).Value!;
        var secondBody = (ChargeResponse)((OkObjectResult)second.Result!).Value!;
        secondBody.Id.Should().Be(firstBody.Id); // idempotent on orderId when header absent

        gateway.Verify(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()), Times.Once);
    }
}
