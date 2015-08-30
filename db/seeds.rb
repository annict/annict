require "csv"

# ==============================================================================
# ChannelGroup
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/channel_groups.csv", headers: true) do |row|
  channel_group = ChannelGroup.create do |cg|
    cg.id          = row[0]
    cg.sc_chgid    = row[1]
    cg.name        = row[2]
    cg.sort_number = row[3]
  end
  puts "channel_group: #{channel_group.id}"
end


# ==============================================================================
# Channel
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/channels.csv", headers: true) do |row|
  channel = Channel.create do |c|
    c.id               = row[0]
    c.channel_group_id = row[1]
    c.sc_chid          = row[2]
    c.name             = row[3]
    c.published        = (row[4] == "true")
  end
  puts "channel: #{channel.id}"
end


# ==============================================================================
# Season
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/seasons.csv", headers: true) do |row|
  season = Season.create do |s|
    s.id   = row[0]
    s.name = row[1]
    s.slug = row[2]
  end
  puts "season: #{season.id}"
end


# ==============================================================================
# Tip
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/tips.csv", headers: true) do |row|
  tip = Tip.create do |t|
    t.id           = row[0]
    t.target       = row[1]
    t.partial_name = row[2]
    t.title        = row[3]
    t.icon_name    = row[4]
  end
  puts "tip: #{tip.id}"
end


# ==============================================================================
# TwitterBot
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/twitter_bots.csv", headers: true) do |row|
  twitter_bot = TwitterBot.create do |tb|
    tb.id   = row[0]
    tb.name = row[1]
  end
  puts "twitter_bot: #{twitter_bot.id}"
end


# ==============================================================================
# Work
# ==============================================================================

limit = ENV.fetch('limit', 50).to_i
works = CSV.read("#{Dir.pwd}/db/data/csv/works.csv", headers: true)
works = limit == 0 ? works.to_a : works.to_a.sample(limit)

works.each do |row|
  work = Work.create do |w|
    w.id                = row[0]
    w.season_id         = row[1].presence || nil
    w.sc_tid            = row[2].presence || nil
    w.title             = row[3]
    w.media             = row[4]
    w.official_site_url = row[5].presence || ""
    w.wikipedia_url     = row[6].presence || ""
    w.watchers_count    = row[8]
    w.released_at       = row[9]
    w.fetch_syobocal    = (row[13] == "true")
    w.twitter_username  = row[14].presence || ""
    w.twitter_hashtag   = row[15].presence || ""
    w.released_at_about = row[16].presence || nil
  end
  puts "work: #{work.id}"
end

work_ids = Work.pluck(:id)


# ==============================================================================
# CoverImage
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/cover_images.csv", headers: true) do |row|
  work_id = row[1].to_i

  if work_ids.include?(work_id)
    cover_image = CoverImage.create do |ci|
      ci.id        = row[0]
      ci.work_id   = work_id
      ci.file_name = row[2]
      ci.location  = row[3]
    end
    puts "cover_image: #{cover_image.id}"
  end
end


# ==============================================================================
# Episode
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/episodes.csv", headers: true) do |row|
  work_id = row[1].to_i

  if work_ids.include?(work_id)
    episode = Episode.create do |e|
      e.id              = row[0]
      e.work_id         = work_id
      e.number          = row[2]
      e.sort_number     = row[3]
      e.sc_count        = row[4]
      e.title           = row[5]
    end
    puts "episode: #{episode.id}"
  end
end


# ==============================================================================
# Program
# ==============================================================================

CSV.foreach("#{Dir.pwd}/db/data/csv/programs.csv", headers: true) do |row|
  work_id = row[3].to_i

  if work_ids.include?(work_id)
    program = Program.create do |p|
      p.id             = row[0]
      p.channel_id     = row[1]
      p.episode_id     = row[2]
      p.work_id        = work_id
      p.started_at     = row[4]
      p.sc_last_update = row[5]
    end
    puts "program: #{program.id}"
  end
end


# ==============================================================================
# Item
# ==============================================================================

identicon = RubyIdenticon.create("1", background_color: "ffffff")
Item.create do |i|
  i.id          = 1
  i.name        = "no name"
  i.url         = "http://example.com"
  i.main        = false
  i.tombo_image = StringIO.new(identicon)
end

CSV.foreach("#{Dir.pwd}/db/data/csv/items.csv", headers: true) do |row|
  work_id = row[1].to_i

  if row[0] != "1" && work_ids.include?(work_id)
    identicon = RubyIdenticon.create(row[0], background_color: "ffffff")
    item = Item.create do |i|
      i.id          = row[0]
      i.work_id     = work_id
      i.name        = row[2]
      i.url         = row[3]
      i.main        = (row[4] == "true")
      i.tombo_image = StringIO.new(identicon)
    end
    puts "item: #{item.id}"
  end
end


# ==============================================================================
# User
# ==============================================================================

10.times do |i|
  user = User.new do |u|
    u.username = "user#{i + 1}"
    u.email = "user#{i + 1}@example.com"
    u.password = "password"
  end

  user.confirm

  oauth = {
    provider: "twitter",
    uid: i + 1,
    credentials: {
      token: "tokentokentokentokentoken",
      expires_at: Time.now,
      secret: "secretsecretsecretsecretsecret"
    },
    info: {
      name: "User #{i + 1}",
      description: "descriptiondescriptiondescriptiondescriptiondescription",
      image: "https://robohash.org/user#{i + 1}.png"
    }
  }
  user.build_relations(oauth)

  user.save!
  puts "user: #{user.id}"

  episode_ids = Episode.pluck(:id).sample(20)
  episode_ids.each do |episode_id|
    episode = Episode.find(episode_id)
    checkin = user.checkins.create(episode_id: episode_id, work_id: episode.work.id)
    puts "checkin: #{checkin.id}"
  end
end
