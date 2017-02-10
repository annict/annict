# frozen_string_literal: true

class AsinOrCopyrightValidator < ActiveModel::Validator
  def validate(record)
    return if record.asin.present? || record.copyright.present?

    message = I18n.t("resources.work_image.asin_or_copyright_required")
    record.errors[:asin_or_copyright] << message
  end
end
