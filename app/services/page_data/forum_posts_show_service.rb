# frozen_string_literal: true

module PageData
  class ForumPostsShowService
    def self.exec(page_params)
      new.exec(page_params)
    end

    def exec(page_params)
      {
        post: ForumPost.find(page_params.dig("post", "id")),
        comments: ForumComment.where(id: page_params["comments"].pluck("id"))
      }
    end
  end
end
