# frozen_string_literal: true

module PageData
  class WorksSeasonService
    def self.exec(current_user, page_params)
      new.exec(current_user, page_params)
    end

    def exec(current_user, page_params)
      work_ids = page_params["works"].pluck("id").uniq
      display_option = page_params["display_option"]

      {
        work_ids: work_ids,
        statuses: Work.statuses(work_ids, current_user),
        users_data: display_option == "list_detailed" ? Work.watching_friends_data(work_ids, current_user) : nil
      }
    end
  end
end
