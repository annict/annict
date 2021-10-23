# frozen_string_literal: true

class Season
  extend Enumerize

  YEAR_LIST = (1900..(Time.now.year + 5))
  NAME_HASH = {winter: 1, spring: 2, summer: 3, autumn: 4}.freeze

  attr_reader :year, :name

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

  def self.list(sort: :asc, include_all: false)
    years = YEAR_LIST.sort
    years = years.reverse if sort == :desc
    name_hash = NAME_HASH
    name_hash = name_hash.merge(all: 0) if include_all
    names = name_hash.sort_by { |_, v| sort == :desc ? -v : v }.to_h.keys

    years.map { |year|
      names.map do |name|
        new(year, name)
      end
    }.flatten
  end

  def self.find_by_slug(slug)
    year, name = slug&.to_s&.split("-")
    year_exists = year.present? && year.to_i.in?(YEAR_LIST)
    name_exists = name.present? && (name == "all" || name.to_sym.in?(NAME_HASH.keys))

    return if !year_exists || !name_exists

    new(year, name)
  end

  def self.current
    find_by_slug(ENV.fetch("ANNICT_CURRENT_SEASON"))
  end

  def self.next
    find_by_slug(ENV.fetch("ANNICT_NEXT_SEASON"))
  end

  def self.prev
    find_by_slug(ENV.fetch("ANNICT_PREVIOUS_SEASON"))
  end

  def self.no_season
    new(nil, nil)
  end

  def initialize(year, name)
    @year = year
    @name = name
  end

  def work_conditions
    {
      season_year: @year,
      season_name: all? ? nil : name_value
    }
  end

  def all?
    @name == "all"
  end

  def no_season?
    @year.nil? && @name.nil?
  end

  def slug
    "#{@year}-#{@name}"
  end

  def name_value
    NAME_HASH[@name.to_sym].presence || 0
  end

  def local_name(locale = nil)
    I18n.with_locale(locale) do
      return I18n.t("resources.season.no_season") if no_season?

      I18n.t("resources.season.yearly.#{@name}", year: @year)
    end
  end

  def sibling_season(position)
    if all?
      sibling_year = case position
      when :prev then @year.to_i - 1
      when :next then @year.to_i + 1
      else
        @year.to_i
      end

      return unless sibling_year.in?(YEAR_LIST)

      Season.new(sibling_year, "all")
    else
      slugs = Season.list.map(&:slug)
      current_index = slugs.index(slug)

      sibling_index = case position
      when :prev then current_index - 1
      when :next then current_index + 1
      else
        current_index
      end
      sibling_slug = slugs[sibling_index]

      return if sibling_slug.blank?

      Season.find_by_slug(sibling_slug)
    end
  end

  def color
    case @name
    when "winter" then "#78909c"
    when "spring" then "#ec407a"
    when "summer" then "#42a5f5"
    when "autumn" then "#ff7043"
    else
      "#66bb6a"
    end
  end

  def icon_name
    case @name
    when "winter" then "snowflakes"
    when "spring" then "flower-daffodil"
    when "summer" then "island-tropical"
    when "autumn" then "pumpkin"
    else
      raise
    end
  end
end
