# frozen_string_literal: true

module AnimeTagsHelper
  def build_work_tags_json(user, work)
    user.tags_by_work(work).map do |work_tag|
      {
        name: work_tag.name
      }
    end.to_json
  end
end
