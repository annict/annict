# frozen_string_literal: true

module OrganizationCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(name name_kana url wikipedia_url twitter_username).freeze
  PUBLISH_FIELDS = DIFF_FIELDS

  included do
    validates :name, presence: true
    validates :name_kana, presence: true
    validates :url, url: { allow_blank: true }
    validates :wikipedia_url, url: { allow_blank: true }

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = send(field)
        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
