// Mongo shell script - the MongoDB equivalent of a schema migration for a
// schemaless document store: it only needs to declare indexes, since the
// document shape itself is enforced by the Mongoose schema in
// src/models/Notification.ts, not by the database.
//
// Run with:
//   mongosh "$MONGO_URI" scripts/init-indexes.js
// or mount it into the official mongo image's /docker-entrypoint-initdb.d/.

db = db.getSiblingDB(db.getName());

db.notifications.createIndex({ userId: 1, createdAt: -1 });
db.notifications.createIndex({ status: 1 });

print('notification-service: indexes created on notifications collection');
