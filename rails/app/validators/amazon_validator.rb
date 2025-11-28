# typed: false

class AmazonValidator < ActiveModel::EachValidator
  def validate_each(record, _attribute, value)
    if value.present?
      begin
        url = URI.parse(value)
        unless /amazon\.co\.jp\z/.match?(url.host)
          record.errors.add(:url, "にはAmazon.co.jpの商品URLを入力してください。")
        end
      rescue URI::InvalidURIError => err
        if err.message.include?("URI must be ascii only")
          record.errors.add(:url, "には日本語を含めることはできません。URLをエンコードするか、日本語を取り除いてください。")
        end
      end
    end
  end
end
