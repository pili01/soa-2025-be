const prisma = require('../config/prisma');

exports.create = async (blogData) => {
  return prisma.blog.create({
    data: blogData,
  });
};

exports.getAll = async () => {
  return prisma.blog.findMany();
};

exports.getBlogs = async (skip, limit, followed) => {
  return await prisma.blog.findMany({
    where: {
      userId: { in: followed }
    },
    skip: skip,
    take: limit
  })
};

exports.getById = async (blogId) => {
  return await prisma.blog.findUnique({
    where: {
      id: blogId
    }
  });
};
