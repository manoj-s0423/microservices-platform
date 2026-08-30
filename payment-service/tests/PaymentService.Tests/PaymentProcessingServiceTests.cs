using FluentAssertions;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging.Abstractions;
using Moq;
using PaymentService.Api.Data;
using PaymentService.Api.Models;
using PaymentService.Api.Services;
using Xunit;

namespace PaymentService.Tests;

public class PaymentProcessingServiceTests
{
    private static PaymentDbContext NewInMemoryContext()
    {
        var options = new DbContextOptionsBuilder<PaymentDbContext>()
            .UseInMemoryDatabase(Guid.NewGuid().ToString())
            .Options;
        return new PaymentDbContext(options);
    }

    private static ChargeRequest ValidRequest(int amountCents = 2000) => new()
    {
        OrderId = Guid.NewGuid().ToString(),
        UserId = Guid.NewGuid().ToString(),
        AmountCents = amountCents,
        Currency = "USD",
    };

    [Fact]
    public async Task ChargeAsync_ApprovedByGateway_ReturnsSucceededPayment()
    {
        var db = NewInMemoryContext();
        var gateway = new Mock<IPaymentGatewayClient>();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(new GatewayChargeResult(true, "txn_abc123", null));

        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);
        var request = ValidRequest();

        var payment = await service.ChargeAsync(request, request.OrderId, CancellationToken.None);

        payment.Status.Should().Be(PaymentStatus.Succeeded);
        payment.TransactionId.Should().Be("txn_abc123");
    }

    [Fact]
    public async Task ChargeAsync_DeclinedByGateway_ReturnsDeclinedPayment()
    {
        var db = NewInMemoryContext();
        var gateway = new Mock<IPaymentGatewayClient>();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(new GatewayChargeResult(false, null, "insufficient_funds"));

        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);
        var request = ValidRequest();

        var payment = await service.ChargeAsync(request, request.OrderId, CancellationToken.None);

        payment.Status.Should().Be(PaymentStatus.Declined);
        payment.FailureReason.Should().Be("insufficient_funds");
    }

    [Fact]
    public async Task ChargeAsync_GatewayUnavailable_ReturnsFailedPayment_NotAnException()
    {
        var db = NewInMemoryContext();
        var gateway = new Mock<IPaymentGatewayClient>();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ThrowsAsync(new PaymentGatewayUnavailableException("boom", new HttpRequestException()));

        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);
        var request = ValidRequest();

        var payment = await service.ChargeAsync(request, request.OrderId, CancellationToken.None);

        payment.Status.Should().Be(PaymentStatus.Failed);
        payment.FailureReason.Should().Be("gateway_unavailable");
    }

    [Fact]
    public async Task ChargeAsync_RepeatedIdempotencyKey_ReturnsStoredResultWithoutCallingGatewayAgain()
    {
        var db = NewInMemoryContext();
        var gateway = new Mock<IPaymentGatewayClient>();
        gateway.Setup(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()))
            .ReturnsAsync(new GatewayChargeResult(true, "txn_first", null));

        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);
        var request = ValidRequest();

        var first = await service.ChargeAsync(request, request.OrderId, CancellationToken.None);
        var second = await service.ChargeAsync(request, request.OrderId, CancellationToken.None);

        second.Id.Should().Be(first.Id);
        second.TransactionId.Should().Be("txn_first");
        gateway.Verify(g => g.ChargeAsync(It.IsAny<GatewayChargeRequest>(), It.IsAny<CancellationToken>()), Times.Once);
    }

    [Fact]
    public async Task ChargeAsync_InvalidOrderId_ThrowsArgumentException()
    {
        var db = NewInMemoryContext();
        var gateway = new Mock<IPaymentGatewayClient>();
        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);

        var request = ValidRequest();
        request.OrderId = "not-a-guid";

        Func<Task> act = () => service.ChargeAsync(request, "some-key", CancellationToken.None);

        await act.Should().ThrowAsync<ArgumentException>();
    }

    [Fact]
    public async Task GetByIdAsync_UnknownId_ReturnsNull()
    {
        var db = NewInMemoryContext();
        var gateway = new Mock<IPaymentGatewayClient>();
        var service = new PaymentProcessingService(db, gateway.Object, NullLogger<PaymentProcessingService>.Instance);

        var result = await service.GetByIdAsync(Guid.NewGuid(), CancellationToken.None);

        result.Should().BeNull();
    }
}
