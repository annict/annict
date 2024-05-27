# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: forum_categories
#
#  id                :bigint           not null, primary key
#  description       :string           not null
#  description_en    :string           not null
#  forum_posts_count :integer          default(0), not null
#  name              :string           not null
#  name_en           :string           not null
#  postable_role     :string           not null
#  slug              :string           not null
#  sort_number       :integer          not null
#  created_at        :datetime         not null
#  updated_at        :datetime         not null
#
# Indexes
#
#  index_forum_categories_on_slug  (slug) UNIQUE
#

class ForumCategory < ApplicationRecord
  extend Enumerize

  enumerize :slug, in: %i[site_news general feedback db_request], scope: true

  has_many :forum_posts, inverse_of: :forum_category, dependent: :destroy

  scope :selectable, ->(user) {
    return all if user.role.admin?
    where.not(slug: :site_news)
  }
end
