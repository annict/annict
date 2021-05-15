# frozen_string_literal: true

module Api
  module Internal
    class WorkFriendsController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?
        return render(json: []) if params[:display_option] != "list_detailed"
        return render(json: []) if work_ids.empty?

        friends_data = Work.watching_friends_data(work_ids, current_user)

        result = work_ids.map { |work_id|
          data = {
            work_id: work_id
          }

          data[:users] = friends_data
            .select { |ud| ud[:work_id] == work_id }
            .first[:users_data]
            .sort_by { |ud| ud[:library_entry_id] }
            .reverse
            .map { |ud|
              {
                username: ud[:user].username,
                avatar_url: helpers.v4_ann_image_url(ud[:user].profile, :image, size: "30x30")
              }
            }

          data
        }

        render json: result
      end

      private

      def work_ids
        ids = params[:work_ids]&.split(",")

        return [] if !ids || !ids.all? { |id| %r{\A[0-9]+\z}.match?(id) }

        ids.map(&:to_i)
      end
    end
  end
end
