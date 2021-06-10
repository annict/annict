# frozen_string_literal: true

class SpoilerGuardComponent < ApplicationComponent
  def initialize(work_id:, episode_id: nil)
    @work_id = work_id
    @episode_id = episode_id
  end

  private

  attr_reader :work_id, :episode_id
end
