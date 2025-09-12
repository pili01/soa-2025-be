-- CreateTable
CREATE TABLE "public"."blogs" (
    "id" SERIAL NOT NULL,
    "userId" INTEGER NOT NULL,
    "title" TEXT NOT NULL,
    "content" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "blogs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "public"."blog_likes" (
    "userId" INTEGER NOT NULL,
    "blogId" INTEGER NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "blog_likes_pkey" PRIMARY KEY ("userId","blogId")
);

-- CreateTable
CREATE TABLE "public"."blog_comments" (
    "id" SERIAL NOT NULL,
    "blogId" INTEGER NOT NULL,
    "userId" INTEGER NOT NULL,
    "author_username" TEXT NOT NULL,
    "content" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "blog_comments_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "blogs_userId_idx" ON "public"."blogs"("userId");

-- CreateIndex
CREATE INDEX "blog_likes_blogId_idx" ON "public"."blog_likes"("blogId");

-- CreateIndex
CREATE INDEX "blog_likes_created_at_idx" ON "public"."blog_likes"("created_at");

-- CreateIndex
CREATE INDEX "blog_comments_blogId_idx" ON "public"."blog_comments"("blogId");

-- CreateIndex
CREATE INDEX "blog_comments_userId_idx" ON "public"."blog_comments"("userId");

-- AddForeignKey
ALTER TABLE "public"."blog_likes" ADD CONSTRAINT "blog_likes_blogId_fkey" FOREIGN KEY ("blogId") REFERENCES "public"."blogs"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."blog_comments" ADD CONSTRAINT "blog_comments_blogId_fkey" FOREIGN KEY ("blogId") REFERENCES "public"."blogs"("id") ON DELETE CASCADE ON UPDATE CASCADE;
