# frozen_string_literal: true

module Forms
  class WorkRecordForm < Forms::ApplicationForm
    attr_accessor :work, :record, :oauth_application,
      :rating, :animation_rating, :character_rating, :music_rating, :story_rating
    attr_reader :body, :deprecated_title, :share_to_twitter

    validates :work, presence: true
    validates :body, length: {maximum: 1_048_596}

    Record::RATING_COLUMNS.each do |rating_column|
      validates rating_column, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}

      define_method "#{rating_column}=" do |value|
        instance_variable_set "@#{rating_column}", value&.downcase.presence
      end
    end

    def deprecated_title=(value)
      @deprecated_title = value&.strip
    end

    def body=(value)
      @body = value&.strip
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
