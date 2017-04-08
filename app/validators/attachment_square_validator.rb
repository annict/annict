# frozen_string_literal: true

class AttachmentSquareValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?

    image_path = Paperclip.io_adapters.for(value).path
    image = MiniMagick::Image.open(image_path)

    return if image.width == image.height

    record.errors.add(attribute, "には縦横1対1の画像を指定してください。")
  end
end
