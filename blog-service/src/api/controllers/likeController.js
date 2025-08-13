const likeService = require('../../service/likeService');

exports.create = async(req, res, next) => {
    try{
        const likeData = req.body;
        const newLike = await likeService.createLike(likeData);

        res.status(201).json({
            success: true,
            message: 'You managed to create new like on blog',
            data: newLike
        });
    }catch(err){
        res.status(500).json({
            success: false,
            message: 'An error occurred on the server.',
        });
    }
}

exports.getAllBlogLikes = async (req, res, next) => {
  try {
    const { blogId } = req.params;
    if (!blogId) {
      return res.status(400).json({ success: false, message: 'Missing blogId in params.' });
    }

    const likes = await likeService.getAllBlogLikes(+blogId);

    return res.status(200).json({
      success: true,
      message: 'Fetched all likes for the blog.',
      data: likes
    });
  } catch (err) {
    return res.status(500).json({
      success: false,
      message: 'An error occurred on the server.'
    });
  }
};

exports.deleteBlogLike = async (req, res, next) => {
  try {
    const { blogId, userId } = req.params;
    if (!blogId || !userId) {
      return res.status(400).json({ success: false, message: 'Missing blogId or userId in params.' });
    }

    const deleted = await likeService.delete(+blogId, +userId);
    if (!deleted) {
      return res.status(404).json({ success: false, message: 'Like not found for the given blogId/userId.' });
    }

    return res.status(204).send();
  } catch (err) {
    return res.status(500).json({ success: false, message: 'An error occurred on the server.' });
  }
};