# frozen_string_literal: true

class ImageUploaderBase < Shrine
  ALLOWED_TYPES = %w[image/gif image/jpeg image/png].freeze
  MAX_SIZE = 10.megabytes
  MAX_SIDE_LENGTH = 5000

  plugin :remove_attachment
  plugin :pretty_location
  plugin :processing
  plugin :versions
  plugin :validation_helpers
  plugin :store_dimensions, analyzer: :mini_magick

  Attacher.validate do
    validate_max_size MAX_SIZE

    if validate_mime_type_inclusion(ALLOWED_TYPES)
      validate_max_width MAX_SIDE_LENGTH
      validate_max_height MAX_SIDE_LENGTH
    end
  end

  process(:store) do |io, _context|
    unless io.original_filename
      ext = MiniMime.lookup_by_content_type(io.metadata["mime_type"]).extension
      io.data["metadata"]["filename"] = "#{io.hash}.#{ext}" if ext
    end

    versions = {original: io}

    io.download do |original|
      versions[:master] = ImageProcessing::MiniMagick
        .source(original)
        .loader(page: 0) # For gif animation image
        .convert(:jpg)
        .saver(quality: 90)
        .strip
        .resize_to_limit(1000, nil, sharpen: false)
        .call
    end

    versions
  end
end
