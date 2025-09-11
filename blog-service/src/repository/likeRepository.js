const prisma = require('../config/prisma')

exports.create = async (likeData) => {
  return prisma.blogLike.create({
    data: {
      ...likeData,
      userId: Number(likeData.userId),
      blogId: Number(likeData.blogId),
    },
  });
};

exports.getAll = async () => {
  return prisma.blogLike.findMany();
};

exports.getAllBlogLikes = async (blogId) => {
  return prisma.blogLike.findMany({
    where: { blogId: Number(blogId) },
  });
};

exports.findByUserAndBlog = (userId, blogId) =>
  prisma.blogLike.findUnique({
    where: {
      userId_blogId: {
        userId: Number(userId),
        blogId: Number(blogId),
      },
    },
  });

exports.countByBlog = (blogId) =>
  prisma.blogLike.count({
    where: { blogId: Number(blogId) },
  });

exports.delete = async (userId, blogId) => {
  try {
    return await prisma.blogLike.delete({
      where: {
        userId_blogId: {
          userId: Number(userId),
          blogId: Number(blogId),
        },
      },
    });
  } catch (err) {
    if (err.code === 'P2025') {
      return null;
    }
    throw err;
  }
};
