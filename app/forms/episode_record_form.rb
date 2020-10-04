# frozen_string_literal: true

class EpisodeRecordForm < ApplicationForm
  attr_accessor :comment, :episode_id, :record_id, :rating_state

  def persisted?
    !record_id.nil?
  end
end
