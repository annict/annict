class AmazonValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    if value.present?
      unless /amazon\.co\.jp\z/ === URI.parse(value).host
        record.errors.add(:url, "にはAmazon.co.jpの商品URLを入力してください。")
      end
    end
  end
end
