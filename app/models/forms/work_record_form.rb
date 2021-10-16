# frozen_string_literal: true

module Forms
  class WorkRecordForm < Forms::ApplicationForm
    include FormRecordable

    attr_accessor :work, :oauth_application,
      :rating_animation, :rating_character, :rating_music, :rating_overall, :rating_story,
      :record
    attr_reader :comment, :deprecated_title, :share_to_twitter

    validates :work, presence: true
    validates :comment, length: {maximum: 1_048_596}

    WorkRecord::RATING_FIELDS.each do |rating_field|
      validates rating_field, allow_nil: true, inclusion: {in: Record::RATING_STATES.map(&:to_s)}

      define_method "#{rating_field}=" do |value|
        instance_variable_set "@#{rating_field}", value&.downcase.presence
      end
    end

    def deprecated_title=(value)
      @deprecated_title = value&.strip
    end

    def comment=(comment)
      @comment = comment&.strip
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
