# frozen_string_literal: true

module PersonCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(prefecture_id name name_kana nickname gender url wikipedia_url
                   twitter_username birthday blood_type height).freeze
  PUBLISH_FIELDS = DIFF_FIELDS

  included do
    enumerize :blood_type, in: [:a, :b, :ab, :o]
    enumerize :gender, in: [:male, :female]

    validates :name, presence: true
    validates :name_kana, presence: true
    validates :url, url: { allow_blank: true }
    validates :wikipedia_url, url: { allow_blank: true }

    def attributes=(params)
      super
      self.birthday = Date.parse(params[:birthday]) if params[:birthday].present?
    end

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = send(field)
        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
