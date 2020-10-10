# frozen_string_literal: true

class EpisodeRecordForm < ApplicationForm
  attr_accessor :comment, :episode_id, :rating, :record_id, :share_to_twitter

  def persisted?
    !record_id.nil?
  end
end
