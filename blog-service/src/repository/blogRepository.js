const prisma = require('../config/prisma');

exports.create = async (blogData) => {
  return prisma.blog.create({
    data: blogData,
  });
};

exports.getAll = async () => {
  return prisma.blog.findMany();
};