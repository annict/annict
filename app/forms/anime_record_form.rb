# frozen_string_literal: true

class AnimeRecordForm < ApplicationForm
  attr_accessor :anime_id, :rating_overall, :rating_animation, :rating_music, :rating_story, :rating_character, :record_id
  attr_reader :comment, :share_to_twitter

  WorkRecord::RATING_KINDS.each do |rating_kind|
    validates rating_kind, inclusion: { in: ApplicationEntity::Types::RecordRating.values }
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
