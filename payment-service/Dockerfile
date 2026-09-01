# payment-service — C# / ASP.NET Core
#
# TWO-STAGE build, same reasoning as order-service: `dotnet publish`
# compiles and links your C# into IL assemblies. The SDK image that does
# that compiling (~800MB — full C# compiler, MSBuild, NuGet client) has
# no purpose once those assemblies exist. The ASP.NET **runtime-only**
# image that executes them ships no compiler at all and is a fraction
# of the size.

# --- Stage 1: build -------------------------------------------------
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build

WORKDIR /src

# Manifest-first, same caching principle as npm ci / go mod download:
# copy only the .csproj for the project this image actually ships —
# PaymentService.Api, NOT the solution file, and NOT PaymentService.Tests
# (that project never runs in production and has no business being
# restored, built, or published into this image). Restore its NuGet
# packages first; as long as this one file doesn't change, that restore
# is reused from cache even when source changes on every commit.
COPY src/PaymentService.Api/PaymentService.Api.csproj src/PaymentService.Api/
RUN dotnet restore src/PaymentService.Api/PaymentService.Api.csproj

COPY src/PaymentService.Api/ src/PaymentService.Api/

# publish, not build: `dotnet publish` produces the Release-optimized,
# ready-to-run output meant to be deployed. `dotnet build` alone is a
# compile step for local development/IDE use, not what ships.
# --no-restore: already restored above, no need to repeat it.
RUN dotnet publish src/PaymentService.Api/PaymentService.Api.csproj \
    -c Release \
    --no-restore \
    -o /app/publish

# --- Stage 2: runtime -------------------------------------------------
FROM mcr.microsoft.com/dotnet/aspnet:8.0

WORKDIR /app

COPY --from=build /app/publish .

# .NET 8's official runtime images ship a built-in non-root "app" user,
# exposed via the $APP_UID environment variable the base image already
# defines — the same idea as node's built-in `node` user. Unlike the
# alpine-based order-service image, there's no addgroup/adduser needed
# here; Microsoft did that work in the base image already.
USER $APP_UID

# ASPNETCORE_URLS is what actually tells Kestrel where to bind — EXPOSE
# alone is documentation, it doesn't make the app listen anywhere.
ENV ASPNETCORE_URLS=http://+:8083
EXPOSE 8083

# NOTE on InvariantGlobalization: PaymentService.Api.csproj already sets
# <InvariantGlobalization>true</InvariantGlobalization>, so the published
# output has no runtime dependency on ICU globalization data — that
# benefit is already baked in by the publish step above; no extra env
# var needed here to get it.

# No HEALTHCHECK here, deliberately: mcr.microsoft.com/dotnet/aspnet is
# Debian-based and ships neither curl nor wget, so a healthcheck command
# would mean installing a package just to make one HTTP call — unlike
# order-service (alpine's wget) or the Node services (node's own http
# module), .NET has no equivalent already sitting in the base image.
# Given this stack's actual target is Kubernetes, where a livenessProbe/
# readinessProbe hitting GET /health supersedes Docker's own HEALTHCHECK
# anyway, that's a better place to define this check than paying for an
# extra package here. Revisit if this image ever runs outside an
# orchestrator that provides its own probing.

# Exec form, entrypoint directly on `dotnet` — matches the AssemblyName
# in PaymentService.Api.csproj. This process is PID 1, so it receives
# SIGTERM directly, which is what ASP.NET Core's own graceful-shutdown
# handling (draining in-flight requests before exit) depends on.
ENTRYPOINT ["dotnet", "PaymentService.Api.dll"]
