const prisma = require('../config/prisma')

exports.create = async(likeData) => {
    return prisma.blogLike.create({
        data: likeData
    });
};

exports.getAll = async () => {
    return prisma.blogLike.getAll();
}

exports.getAllBlogLikes = async (blogId) => {
  return prisma.blogLike.findMany({
    where: { blogId }
  });
};

exports.delete = async (userId, blogId) => {
  try {
    return await prisma.blogLike.delete({
      where: {
        userId_blogId: {
          userId,
          blogId
        }
      }
    });
  } catch (err) {
    if (err.code === 'P2025') {
      return null;
    }
    throw err;
  }
};