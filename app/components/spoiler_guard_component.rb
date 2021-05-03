# frozen_string_literal: true

# TODO: SpoilerGuardComponent2 に置き換える
class SpoilerGuardComponent < ApplicationComponent
  def initialize(work_id:, episode_id: nil)
    @work_id = work_id
    @episode_id = episode_id
  end

  private

  attr_reader :work_id, :episode_id
end
