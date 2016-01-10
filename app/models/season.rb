# == Schema Information
#
# Table name: seasons
#
#  id          :integer          not null, primary key
#  name        :string(510)      not null
#  created_at  :datetime
#  updated_at  :datetime
#  sort_number :integer          not null
#  year        :integer          not null
#
# Indexes
#
#  index_seasons_on_sort_number    (sort_number) UNIQUE
#  index_seasons_on_year           (year)
#  index_seasons_on_year_and_name  (year,name) UNIQUE
#

class Season < ActiveRecord::Base
  has_many :works

  delegate :yearly_season_ja, to: :decorate

  NAME_DATA = {
    winter: "冬",
    spring: "春",
    summer: "夏",
    autumn: "秋"
  }.freeze

  def self.find_or_new_by_slug(slug)
    year, name = slug.split("-")
    attrs = { year: year, name: name }

    # seasonsテーブルには name: "all" なレコードを保存していないため、newする
    return new(attrs) if name == "all"

    find_by(attrs)
  end

  def self.slug_options
    options = []
    years.each do |year|
      options << ["#{year}年", "#{year}-all"]
      NAME_DATA.reverse_each do |key, val|
        options << ["#{year}年#{val}", "#{year}-#{key}"]
      end
    end
    options
  end

  def self.years
    # 降順で取得する
    pluck(:year).uniq.sort { |a, b| b <=> a }
  end

  def slug
    "#{year}-#{name}"
  end
end
