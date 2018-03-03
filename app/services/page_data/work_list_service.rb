# frozen_string_literal: true

module PageData
  class WorkListService
    def self.exec(current_user, page_params)
      new.exec(current_user, page_params)
    end

    def exec(current_user, page_params)
      @current_user = current_user
      @work_ids = page_params["works"].pluck("id").uniq
      @display_option = page_params["display_option"]

      {
        statuses: Work.statuses(@work_ids, current_user),
        users_data: users_data
      }
    end

    private

    def users_data
      friends_data = Work.watching_friends_data(@work_ids, @current_user) if @display_option == "list_detailed"

      @work_ids.map do |work_id|
        data = {
          work_id: work_id
        }

        next if @display_option != "list_detailed"

        data[:users] = friends_data.
          select { |ud| ud[:work_id] == work_id }.
          first[:users_data].
          sort_by { |ud| ud[:latest_status_id] }.
          reverse.
          map { |ud| ud[:user] }

        data
      end
    end
  end
end
