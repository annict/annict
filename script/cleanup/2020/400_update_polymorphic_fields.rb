# frozen_string_literal: true

[
  [ActivityGroup, :itemable_type, "WorkRecord", "AnimeRecord"],
  [Activity, :trackable_type, "WorkRecord", "AnimeRecord"],
  [DbActivity, :trackable_type, "Work", "Anime"],
  [DbActivity, :trackable_type, "SeriesWork", "SeriesAnime"],
  [DbActivity, :root_resource_type, "Work", "Anime"],
  [DbActivity, :action, "works.create", "anime.create"],
  [DbActivity, :action, "works.update", "anime.update"],
  [DbActivity, :action, "series_works.create", "series_anime.create"],
  [DbActivity, :action, "series_works.update", "series_anime.update"],
  [Like, :recipient_type, "WorkRecord", "AnimeRecord"],
].each do |(model, target_field, from, to)|
  puts "Running #{model}.where(#{target_field} => #{from}).update_all(#{target_field} => #{to})..."
  model.where(target_field => from).update_all(target_field => to)
end
