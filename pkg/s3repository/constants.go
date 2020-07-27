package s3repository

const MinMultiPartSize = 5 * 1024 * 1024
const MaxParts = 10000                 // Maximum number of parts we will use
const MaxBufferSize = 10 * 1024 * 1024 // Let's not get too crazy with buffers
const SegmentSizeLimit = MaxParts * MaxBufferSize
