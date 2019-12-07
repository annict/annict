# frozen_string_literal: true

class TrackableService
  def initialize(user)
    @user = user
  end

  def library_entries
    LibraryEntry.refresh_next_episode(@user)

    @user.
      library_entries.
      includes(:next_episode, work: :work_image).
      watching.
      has_next_episode.
      order(:position)
  end
end
