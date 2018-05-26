# frozen_string_literal: true

module PageData
  class RecordsIndexService
    def self.exec(current_user, page_params)
      new.exec(current_user, page_params)
    end

    def exec(current_user, page_params)
      work_ids = page_params["works"].pluck("id").uniq

      {
        statuses: Work.statuses(work_ids, current_user)
      }
    end
  end
end
