require 'csv'

# ==============================================================================
# seasons
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/seasons.csv") do |row|
  season = Season.create(name: row[1], slug: row[2])
  puts "created. season name: #{season.name}"
end


# ==============================================================================
# works
# ==============================================================================

File.foreach("#{Dir.pwd}/db/data/csv/works.csv") do |file_row|
  row = CSV.parse(file_row).first

  work = Work.create do |w|
    w.season_id         = ('NULL' == row[1]) ? nil : row[1]
    w.title             = row[3]
    w.media             = row[4]
    w.official_site_url = row[5].presence || ''
    w.wikipedia_url     = row[6].presence || ''
    w.episodes_count    = row[7]
    w.released_at       = row[8]
  end
  puts "created. work title: #{work.title}"

  item = work.items.create do |i|
    i.name = "Work #{work.id}"
    i.url = "http://example.com/work_#{work.id}"
    i.image = File.open("#{Rails.root}/db/data/image/work_#{work.id}.png")
  end
  puts "created. item name: #{item.name}"
end


# ==============================================================================
# episodes
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/episodes.csv") do |row|
  episode = Episode.create(work_id: row[1], number: row[2], sort_number: row[3], title: row[6])
  puts "created. episode number: #{episode.number}, episode title: #{episode.title}"
end


# ==============================================================================
# users
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/users.csv") do |row|
  user = User.create do |u|
    u.username             = row[1]
    u.email                = row[2]
    u.confirmed_at         = row[14]
    u.confirmation_sent_at = row[15]
  end
  puts "created. user username: #{user.username}"
end


# ==============================================================================
# profiles
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/profiles.csv") do |row|
  image_file = File.open("#{Rails.root}/db/data/image/user_#{row[1]}.png")
  profile = Profile.create do |p|
    p.user_id          = row[1]
    p.name             = row[2]
    p.description      = row[5]
    p.avatar           = image_file
    p.background_image = image_file
  end
  puts "created. profile name: #{profile.name}"
end


# ==============================================================================
# providers
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/providers.csv") do |row|
  provider = Provider.create do |p|
    p.user_id          = row[1]
    p.name             = row[2]
    p.uid              = row[3]
    p.token            = row[4]
    p.token_expires_at = row[5]
    p.token_secret     = row[6]
  end
  puts "created. provider name: #{provider.name}"
end
