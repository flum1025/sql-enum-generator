CREATE TABLE "public"."products" (
  "id" SERIAL NOT NULL,
  "name" CHARACTER VARYING (32) NOT NULL,
  PRIMARY KEY ("id")
);

CREATE TABLE "public"."menus" (
  "id" CHARACTER VARYING (8) NOT NULL,
  "name" CHARACTER VARYING (32) NOT NULL,
  PRIMARY KEY ("id")
);

CREATE TABLE "public"."users" (
  "id" UUID NOT NULL DEFAULT gen_random_uuid(),
  "product_id" INTEGER NOT NULL,
  PRIMARY KEY ("id")
);
