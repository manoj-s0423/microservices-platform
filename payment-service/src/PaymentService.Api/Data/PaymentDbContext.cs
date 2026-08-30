using Microsoft.EntityFrameworkCore;
using PaymentService.Api.Models;

namespace PaymentService.Api.Data;

/// <summary>
/// Maps to a schema created by the SQL scripts in /migrations (owned
/// exclusively by payment-service). No other service reads or writes the
/// "payments" table directly.
/// </summary>
public class PaymentDbContext : DbContext
{
    public PaymentDbContext(DbContextOptions<PaymentDbContext> options) : base(options)
    {
    }

    public DbSet<Payment> Payments => Set<Payment>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Payment>(entity =>
        {
            entity.ToTable("payments");
            entity.HasKey(p => p.Id);

            entity.Property(p => p.Id).HasColumnName("id");
            entity.Property(p => p.OrderId).HasColumnName("order_id").IsRequired();
            entity.Property(p => p.UserId).HasColumnName("user_id").IsRequired();
            entity.Property(p => p.AmountCents).HasColumnName("amount_cents").IsRequired();
            entity.Property(p => p.Currency).HasColumnName("currency").HasMaxLength(3).IsRequired();
            entity.Property(p => p.Status).HasColumnName("status")
                .HasConversion<string>().HasMaxLength(20).IsRequired();
            entity.Property(p => p.TransactionId).HasColumnName("transaction_id").HasMaxLength(100);
            entity.Property(p => p.FailureReason).HasColumnName("failure_reason").HasMaxLength(255);
            entity.Property(p => p.IdempotencyKey).HasColumnName("idempotency_key").HasMaxLength(100).IsRequired();
            entity.Property(p => p.CreatedAt).HasColumnName("created_at").IsRequired();
            entity.Property(p => p.UpdatedAt).HasColumnName("updated_at").IsRequired();

            entity.HasIndex(p => p.IdempotencyKey).IsUnique();
            entity.HasIndex(p => p.OrderId);
            entity.HasIndex(p => p.UserId);
        });
    }
}
