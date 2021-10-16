# frozen_string_literal: true

module Forms
  class EpisodeRecordForm < Forms::ApplicationForm
    include FormRecordable

    attr_accessor :deprecated_rating, :episode, :oauth_application, :record
    attr_reader :comment, :rating, :share_to_twitter

    validates :comment, length: {maximum: 1_048_596}
    validates :episode, presence: true
    validates :rating, inclusion: {in: Record::RATING_STATES.map(&:to_s)}, allow_nil: true

    def comment=(comment)
      @comment = comment&.strip
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
