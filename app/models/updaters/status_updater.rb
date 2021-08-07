# frozen_string_literal: true

module Updaters
  class StatusUpdater
    def initialize(user:, form:)
      @user = user
      @form = form
      @anime = @form.anime
    end

    def call
      library_entry = @user.library_entries.find_by(anime: @anime)

      if @form.no_status?
        ActiveRecord::Base.transaction do
          library_entry&.update!(status_id: nil)
          UserWatchedWorksCountJob.perform_later(@user.id)
        end
      else
        v2_kind = Status.kind_v3_to_v2(@form.kind)
        status = @user.statuses.new(anime: @anime, kind: v2_kind)

        ActiveRecord::Base.transaction do
          status.save!
          status.save_library_entry

          activity_group = @user.create_or_last_activity_group!(status)
          @user.activities.create!(itemable: status, activity_group: activity_group)

          prev_state_kind = library_entry&.status&.kind
          @user.update_works_count!(prev_state_kind, v2_kind)
          @anime.update_watchers_count!(prev_state_kind, v2_kind)

          UserWatchedWorksCountJob.perform_later(@user.id)
          status.share_to_sns
        end
      end
    end
  end
end
