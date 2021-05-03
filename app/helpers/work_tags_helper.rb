# frozen_string_literal: true

module WorkTagsHelper
  def build_work_tags_json(user, work)
    user.tags_by_work(work).map { |work_tag|
      {
        name: work_tag.name
      }
    }.to_json
  end
end
