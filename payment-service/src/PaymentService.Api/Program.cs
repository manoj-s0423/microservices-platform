using Microsoft.EntityFrameworkCore;
using PaymentService.Api.Data;
using PaymentService.Api.Services;
using Polly;
using Polly.Extensions.Http;
using Serilog;
using Serilog.Formatting.Compact;

var builder = WebApplication.CreateBuilder(args);

// --- Structured (JSON) logging ---
builder.Host.UseSerilog((context, services, config) =>
{
    config
        .Enrich.WithProperty("service", "payment-service")
        .Enrich.FromLogContext()
        .WriteTo.Console(new RenderedCompactJsonFormatter())
        .ReadFrom.Configuration(context.Configuration);
});

// --- Configuration ---
var connectionString = BuildConnectionString(builder.Configuration);
builder.Services.Configure<GatewayOptions>(builder.Configuration.GetSection(GatewayOptions.SectionName));

// --- Graceful shutdown: let in-flight requests finish before the process exits ---
builder.Services.Configure<HostOptions>(opts =>
{
    var seconds = builder.Configuration.GetValue<int?>("SHUTDOWN_TIMEOUT_SECONDS") ?? 10;
    opts.ShutdownTimeout = TimeSpan.FromSeconds(seconds);
});

// --- Database ---
builder.Services.AddDbContext<PaymentDbContext>(options =>
    options.UseNpgsql(connectionString, npgsql => npgsql.CommandTimeout(5)));

// --- External payment gateway client ---
var gatewayOptions = builder.Configuration.GetSection(GatewayOptions.SectionName).Get<GatewayOptions>() ?? new GatewayOptions();

if (string.Equals(gatewayOptions.Mode, "live", StringComparison.OrdinalIgnoreCase))
{
    builder.Services.AddHttpClient<IPaymentGatewayClient, HttpPaymentGatewayClient>(client =>
        {
            client.Timeout = TimeSpan.FromSeconds(gatewayOptions.TimeoutSeconds);
        })
        .AddPolicyHandler(GetRetryPolicy(gatewayOptions.RetryAttempts));
}
else
{
    builder.Services.AddSingleton<IPaymentGatewayClient, SimulatedPaymentGatewayClient>();
}

builder.Services.AddScoped<PaymentProcessingService>();

// --- Health checks ---
builder.Services.AddHealthChecks()
    .AddNpgSql(connectionString, name: "database", tags: new[] { "ready" });

builder.Services.AddControllers();

var app = builder.Build();

app.UseSerilogRequestLogging(); // logs method/path/status/duration with the request's correlation id

app.MapControllers();

// Liveness: process is up. No dependency checks (tags: none matches nothing -> Predicate returns false for all).
app.MapHealthChecks("/health", new Microsoft.AspNetCore.Diagnostics.HealthChecks.HealthCheckOptions
{
    Predicate = _ => false,
});

// Readiness: includes the "ready"-tagged database check.
app.MapHealthChecks("/ready", new Microsoft.AspNetCore.Diagnostics.HealthChecks.HealthCheckOptions
{
    Predicate = check => check.Tags.Contains("ready"),
});

app.Run();

static string BuildConnectionString(IConfiguration configuration)
{
    var host = configuration["DB_HOST"] ?? "localhost";
    var port = configuration["DB_PORT"] ?? "5432";
    var name = configuration["DB_NAME"] ?? "shopstream_payments";
    var user = configuration["DB_USER"] ?? "shopstream";
    var password = configuration["DB_PASSWORD"] ?? "";
    var timeout = configuration["DB_CONNECT_TIMEOUT_SECONDS"] ?? "5";

    return $"Host={host};Port={port};Database={name};Username={user};Password={password};Timeout={timeout}";
}

static IAsyncPolicy<HttpResponseMessage> GetRetryPolicy(int retryAttempts)
{
    return HttpPolicyExtensions
        .HandleTransientHttpError() // 5xx and 408
        .WaitAndRetryAsync(retryAttempts, attempt => TimeSpan.FromMilliseconds(200 * Math.Pow(2, attempt - 1)));
}

// Exposed for WebApplicationFactory-based integration tests.
public partial class Program { }
