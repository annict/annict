# frozen_string_literal: true

class Season
  extend Enumerize

  YEAR_LIST = 1900..(Time.now.year + 2)
  NAME_HASH = { winter: 1, spring: 2, summer: 3, autumn: 4 }.freeze

  enumerize :name, in: NAME_HASH

  def self.latest_slugs
    next_season_slug = ENV.fetch("ANNICT_NEXT_SEASON")
    year = next_season_slug.split("-").first
    years = year.to_i.downto(2000).to_a
    names = Season::NAME_HASH.keys.map(&:to_s).reverse
    slugs = years.map { |y| names.map { |sn| "#{y}-#{sn}" } }.flatten
    index = slugs.index(next_season_slug)
    slugs[index..index + 4]
  end

  def initialize(year, name)
    @year = year
    @name = name
  end

  def work_conditions
    conds = { season_year: @year }
    conds[:season_name] = name.value unless all?
    conds
  end

  def all?
    @name == "all"
  end

  def slug
    "#{@year}-#{@name}"
  end

  def color
    case @name
    when "winter" then "#78909c"
    when "spring" then "#ec407a"
    when "summer" then "#42a5f5"
    when "autumn" then "#ff7043"
    end
  end
end
