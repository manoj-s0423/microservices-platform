// Seed data for local development. References the seeded user from
// user-service (22222222-2222-2222-2222-222222222222) and the seeded
// order from order-service (bbbbbbbb-0000-0000-0000-000000000001).
//
// Run with: mongosh "$MONGO_URI" scripts/seed-data.js

db = db.getSiblingDB(db.getName());

db.notifications.insertMany([
  {
    userId: '22222222-2222-2222-2222-222222222222',
    channel: 'EMAIL',
    type: 'ORDER_CONFIRMED',
    recipient: 'jane.doe@shopstream.dev',
    subject: 'Your ShopStream order is confirmed',
    body: 'Thanks for your order! Order bbbbbbbb-0000-0000-0000-000000000001 has been confirmed.',
    status: 'SENT',
    providerMessageId: 'sim_msg_seed_0001',
    attempts: 1,
    createdAt: new Date(),
    updatedAt: new Date(),
  },
  {
    userId: '22222222-2222-2222-2222-222222222222',
    channel: 'EMAIL',
    type: 'PAYMENT_RECEIPT',
    recipient: 'jane.doe@shopstream.dev',
    subject: 'Receipt for your ShopStream payment',
    body: 'You were charged $39.98 for order bbbbbbbb-0000-0000-0000-000000000001.',
    status: 'SENT',
    providerMessageId: 'sim_msg_seed_0002',
    attempts: 1,
    createdAt: new Date(),
    updatedAt: new Date(),
  },
]);

print('notification-service: seeded 2 sample notifications');
