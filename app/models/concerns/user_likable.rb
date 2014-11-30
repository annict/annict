module UserLikable
  extend ActiveSupport::Concern

  included do
    def like_r?(recipient)
      r_likes.where(recipient: recipient).present?
    end

    #「Recommendable」の `like` メソッドと衝突したため、"_r" というサフィックスをつける羽目になった
    def like_r(recipient)
      r_likes.create(recipient: recipient) unless like_r?(recipient)
    end

    def unlike_r(recipient)
      like = r_likes.where(recipient: recipient).first

      like.destroy if like.present?
    end
  end
end
