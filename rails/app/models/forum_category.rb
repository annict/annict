# typed: false
# frozen_string_literal: true

class ForumCategory < ApplicationRecord
  extend Enumerize

  enumerize :slug, in: %i[site_news general feedback db_request], scope: true

  has_many :forum_posts, inverse_of: :forum_category, dependent: :destroy

  scope :selectable, ->(user) {
    return all if user.role.admin?
    where.not(slug: :site_news)
  }
end
