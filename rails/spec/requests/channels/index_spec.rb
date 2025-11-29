# typed: false
# frozen_string_literal: true

RSpec.describe "GET /channels", type: :request do
  it "チャンネルグループとチャンネルが存在しないとき、正常にアクセスできること" do
    get "/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include("チャンネル")
  end

  it "チャンネルグループとチャンネルが存在するとき、正常に表示されること" do
    channel_group = ChannelGroup.create!(
      name: "地上波",
      sort_number: 1
    )
    Channel.create!(
      channel_group:,
      name: "テレビ東京",
      sort_number: 1
    )

    get "/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include("地上波")
    expect(response.body).to include("テレビ東京")
  end

  it "複数のチャンネルグループとチャンネルが存在するとき、全て表示されること" do
    channel_group1 = ChannelGroup.create!(
      name: "地上波",
      sort_number: 1
    )
    channel_group2 = ChannelGroup.create!(
      name: "BS",
      sort_number: 2
    )
    channel_group3 = ChannelGroup.create!(
      name: "動画配信",
      sort_number: 3
    )

    Channel.create!(
      channel_group: channel_group1,
      name: "フジテレビ",
      sort_number: 1
    )
    Channel.create!(
      channel_group: channel_group1,
      name: "テレビ東京",
      sort_number: 2
    )
    Channel.create!(
      channel_group: channel_group2,
      name: "BS11",
      sort_number: 1
    )
    Channel.create!(
      channel_group: channel_group3,
      name: "dアニメストア",
      sort_number: 1
    )
    Channel.create!(
      channel_group: channel_group3,
      name: "Netflix",
      sort_number: 2
    )

    get "/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include("地上波")
    expect(response.body).to include("フジテレビ")
    expect(response.body).to include("テレビ東京")
    expect(response.body).to include("BS")
    expect(response.body).to include("BS11")
    expect(response.body).to include("動画配信")
    expect(response.body).to include("dアニメストア")
    expect(response.body).to include("Netflix")
  end

  it "unpublishedなチャンネルグループは表示されないこと" do
    channel_group_published = ChannelGroup.create!(
      name: "地上波",
      sort_number: 1
    )
    channel_group_unpublished = ChannelGroup.create!(
      name: "未公開グループ",
      sort_number: 2,
      unpublished_at: Time.current
    )

    Channel.create!(
      channel_group: channel_group_published,
      name: "フジテレビ",
      sort_number: 1
    )
    Channel.create!(
      channel_group: channel_group_unpublished,
      name: "未公開チャンネル",
      sort_number: 1
    )

    get "/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include("地上波")
    expect(response.body).to include("フジテレビ")
    expect(response.body).not_to include("未公開グループ")
    expect(response.body).not_to include("未公開チャンネル")
  end

  it "unpublishedなチャンネルは表示されないこと" do
    channel_group = ChannelGroup.create!(
      name: "地上波",
      sort_number: 1
    )

    Channel.create!(
      channel_group:,
      name: "フジテレビ",
      sort_number: 1
    )
    Channel.create!(
      channel_group:,
      name: "未公開チャンネル",
      sort_number: 2,
      unpublished_at: Time.current
    )

    get "/channels"

    expect(response.status).to eq(200)
    expect(response.body).to include("地上波")
    expect(response.body).to include("フジテレビ")
    expect(response.body).not_to include("未公開チャンネル")
  end

  it "チャンネルグループとチャンネルがソート順に表示されること" do
    channel_group3 = ChannelGroup.create!(
      name: "動画配信",
      sort_number: 3
    )
    channel_group1 = ChannelGroup.create!(
      name: "地上波",
      sort_number: 1
    )
    ChannelGroup.create!(
      name: "BS",
      sort_number: 2
    )

    Channel.create!(
      channel_group: channel_group1,
      name: "テレビ東京",
      sort_number: 2
    )
    Channel.create!(
      channel_group: channel_group1,
      name: "フジテレビ",
      sort_number: 1
    )
    Channel.create!(
      channel_group: channel_group3,
      name: "Netflix",
      sort_number: 2
    )
    Channel.create!(
      channel_group: channel_group3,
      name: "dアニメストア",
      sort_number: 1
    )

    get "/channels"

    expect(response.status).to eq(200)
    body = response.body

    # チャンネルグループの表示順を確認
    group1_pos = body.index("地上波")
    group2_pos = body.index("BS")
    group3_pos = body.index("動画配信")
    expect(group1_pos).to be < group2_pos
    expect(group2_pos).to be < group3_pos

    # 同じグループ内のチャンネルの表示順を確認
    fuji_pos = body.index("フジテレビ")
    tokyo_pos = body.index("テレビ東京")
    expect(fuji_pos).to be < tokyo_pos

    danime_pos = body.index("dアニメストア")
    netflix_pos = body.index("Netflix")
    expect(danime_pos).to be < netflix_pos
  end
end
