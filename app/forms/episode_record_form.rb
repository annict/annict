# frozen_string_literal: true

class EpisodeRecordForm < ApplicationForm
  attr_accessor :episode_id, :rating, :record_id
  attr_reader :comment, :share_to_twitter

  validates :rating, inclusion: { in: ApplicationEntity::Types::RecordRatingState.values }

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
