# typed: false
# frozen_string_literal: true

class TwitterHashtagLinkComponent < ApplicationComponent
  def initialize(entity:, title: nil)
    @entity = entity
    @title = title
  end

  def call
    return "-" if entity.twitter_hashtag.blank?

    link_to link_title, entity.twitter_hashtag_url, target: "_blank", rel: "noopener"
  end

  private

  attr_reader :title, :entity

  def link_title
    title.presence || "##{entity.twitter_hashtag}"
  end
end
