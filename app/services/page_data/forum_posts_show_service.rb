# frozen_string_literal: true

module PageData
  class ForumPostsShowService
    def self.exec(current_user, page_params)
      new.exec(current_user, page_params)
    end

    def exec(current_user, page_params)
      {
        post: ForumPost.find(page_params.dig("post", "id")),
        comments: ForumComment.where(id: page_params["comments"].pluck("id"))
      }
    end
  end
end
