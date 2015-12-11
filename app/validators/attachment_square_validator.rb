class AttachmentSquareValidator < ActiveModel::EachValidator
  def validate_each(record, _attribute, value)
    if value.present?
      image_path = Paperclip.io_adapters.for(value).path
      image = MiniMagick::Image.open(image_path)

      if image.width != image.height
        record.errors.add(:tombo_image, "には縦横1:1の画像を指定してください。")
      end
    end
  end
end
