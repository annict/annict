# frozen_string_literal: true

class AnimeRecordForm < ApplicationForm
  attr_accessor :anime_id, :rating_overall, :rating_animation, :rating_music, :rating_story, :rating_character, :record_id
  attr_reader :comment, :share_to_twitter

  WorkRecord::RATING_FIELDS.each do |rating_field|
    validates rating_field, allow_nil: true, inclusion: { in: ApplicationEntity::Types::RecordRatingState.values }

    define_method "#{rating_field}=" do |value|
      instance_variable_set "@#{rating_field}", value.presence
    end
  end

  def comment=(comment)
    @comment = comment&.strip
  end

  def share_to_twitter=(share_to_twitter)
    @share_to_twitter = ActiveRecord::Type::Boolean.new.serialize(share_to_twitter)
  end

  def persisted?
    !record_id.nil?
  end

  def unique_id
    SecureRandom.uuid
  end
end
