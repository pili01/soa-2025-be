const express = require('express');
const multer = require('multer');
const fs = require('fs');
const path = require('path');
const { v4: uuidv4 } = require('uuid');

const router = express.Router();
const uploadDir = path.join(__dirname, '../../uploads/ProfilePictures');
const reviewUploadDir = path.join(__dirname, '../../uploads/TourReviewPictures');
const keypointUploadDir = path.join(__dirname, '../../uploads/KeypointPictures');
//const uploadDir = path.join(__dirname, '../../uploads/images');

if (!fs.existsSync(uploadDir)) fs.mkdirSync(uploadDir, { recursive: true });
if (!fs.existsSync(reviewUploadDir)) fs.mkdirSync(reviewUploadDir, { recursive: true });
if (!fs.existsSync(keypointUploadDir)) fs.mkdirSync(keypointUploadDir, { recursive: true });

const upload = multer({ storage: multer.memoryStorage() });

router.post('/save-image', upload.single('image'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const uniqueId = uuidv4();
    const ext = path.extname(req.file.originalname);
    const filename = `${uniqueId}-${Date.now()}${ext}`;
    const filePath = path.join(uploadDir, filename);

    fs.writeFileSync(filePath, req.file.buffer);

    res.status(201).json({
        message: 'Image saved successfully',
        photoURL: `${filename}`,
        photoName: `${filename}`
    });
});

router.get('/img/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(uploadDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).json({ error: 'File not found' });
    }

    res.sendFile(filePath);
});


router.post('/saveReviewPhoto', upload.single('image'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const userId = req.body.userId || 'unknown_user';
    const tourId = req.body.tourId || 'unknown_tour';
    const ext = path.extname(req.file.originalname);

    const filename = `tour-${tourId}-user-${userId}-${Date.now()}${ext}`;
    const filePath = path.join(reviewUploadDir, filename);

    fs.writeFileSync(filePath, req.file.buffer);

    res.status(201).json({
        message: 'Review image saved successfully',
        photoURL: `http://localhost:3031/api/images/review/${filename}`,
        photoName: filename
    });
});

router.get('/images/review/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(reviewUploadDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).json({ error: 'File not found' });
    }
    
    res.sendFile(filePath);
});

// Keypoint image upload endpoint
router.post('/saveKeypointPhoto', upload.single('image'), (req, res) => {
    if (!req.file) {
        return res.status(400).json({ error: 'No file uploaded' });
    }

    const tourId = req.body.tourId || 'unknown_tour';
    const keypointId = req.body.keypointId || 'unknown_keypoint';
    const ext = path.extname(req.file.originalname);

    const filename = `tour-${tourId}-keypoint-${keypointId}-${Date.now()}${ext}`;
    const filePath = path.join(keypointUploadDir, filename);

    fs.writeFileSync(filePath, req.file.buffer);

    res.status(201).json({
        message: 'Keypoint image saved successfully',
        photoURL: `http://localhost:3001/api/img/keypoint/${filename}`,
        photoName: filename
    });
});

// Keypoint image retrieval endpoint
router.get('/img/keypoint/:filename', (req, res) => {
    const filename = req.params.filename;
    const filePath = path.join(keypointUploadDir, filename);

    if (!fs.existsSync(filePath)) {
        return res.status(404).json({ error: 'File not found' });
    }
    
    res.sendFile(filePath);
});

module.exports = router;