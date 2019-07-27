# frozen_string_literal: true

module ShareButtonHelper
  def twitter_share_button
    tag :span, class: "c-share-button-twitter" do
      tag :span, class: "btn btn-sm u-btn-twitter" do
        tag :div, class: "small" do
          icon "twitter", "fab", class: "mr-1"
          t "noun.tweet"
        end
      end
    end
  end
end
