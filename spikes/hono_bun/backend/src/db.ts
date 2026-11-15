import postgres from "postgres";

const databaseUrl = Bun.env.DATABASE_URL ?? process.env.DATABASE_URL;

export const sql = databaseUrl ? postgres(databaseUrl) : null;
