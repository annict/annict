# frozen_string_literal: true

module PageData
  class WorkDetailService
    def self.exec(current_user, page_params)
      new.exec(current_user, page_params)
    end

    def exec(current_user, page_params)
      work_id = page_params.dig("work", "id")

      {
        statuses: Work.statuses([work_id], current_user)
      }
    end
  end
end
