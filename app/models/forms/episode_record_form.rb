# frozen_string_literal: true

module Forms
  class EpisodeRecordForm < Forms::ApplicationForm
    attr_accessor :deprecated_rating, :episode, :oauth_application, :record
    attr_reader :body, :rating, :share_to_twitter

    validates :body, length: {maximum: 1_048_596}
    validates :episode, presence: true
    validates :rating, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}, allow_nil: true

    def body=(value)
      @body = value&.strip
    end

    def rating=(rating)
      @rating = rating&.downcase.presence
    end

    def share_to_twitter=(value)
      @share_to_twitter = ActiveModel::Type::Boolean.new.cast(value)
    end

    # @overload
    def persisted?
      !record.nil?
    end

    def unique_id
      @unique_id ||= SecureRandom.uuid
    end
  end
end
