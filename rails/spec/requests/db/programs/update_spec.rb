# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/programs/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトし、プログラムは更新されないこと" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    program = create(:program)
    old_program = program.attributes
    program_params = {
      channel_id: channel.id
    }

    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(program.channel_id).to eq(old_program["channel_id"])
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセス拒否され、プログラムは更新されないこと" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user)
    program = create(:program)
    old_program = program.attributes
    program_params = {
      channel_id: channel.id
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(program.channel_id).to eq(old_program["channel_id"])
  end

  it "編集者権限を持つユーザーがログインしているとき、プログラムが正常に更新されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel = Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program)
    old_program = program.attributes
    program_params = {
      channel_id: channel.id
    }

    login_as(user, scope: :user)
    expect(program.channel_id).to eq(old_program["channel_id"])

    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(program.channel_id).to eq(channel.id)
  end

  it "編集者権限を持つユーザーがログインしているとき、started_atが正常に更新されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program)
    new_started_at = 1.day.from_now
    program_params = {
      started_at: new_started_at
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(program.started_at.to_date).to eq(new_started_at.to_date)
  end

  it "編集者権限を持つユーザーがログインしているとき、rebroadcastが正常に更新されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program, rebroadcast: false)
    program_params = {
      rebroadcast: true
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(program.rebroadcast).to eq(true)
  end

  it "編集者権限を持つユーザーがログインしているとき、vod_title_codeが正常に更新されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program)
    program_params = {
      vod_title_code: "test_code_123"
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(program.vod_title_code).to eq("test_code_123")
  end

  it "編集者権限を持つユーザーがログインしているとき、vod_title_nameが正常に更新されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program)
    program_params = {
      vod_title_name: "テストタイトル"
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(program.vod_title_name).to eq("テストタイトル")
  end

  it "編集者権限を持つユーザーがログインしているとき、minimum_episode_generatable_numberが正常に更新されること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program)
    program_params = {
      minimum_episode_generatable_number: 5
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}
    program.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(program.minimum_episode_generatable_number).to eq(5)
  end

  it "編集者権限を持つユーザーがログインしているとき、不正なパラメータで更新エラーが発生すること" do
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    Channel.create!(channel_group:, name: "テレビ東京", sort_number: 1)
    user = create(:registered_user, :with_editor_role)
    program = create(:program)
    program_params = {
      channel_id: nil
    }

    login_as(user, scope: :user)
    patch "/db/programs/#{program.id}", params: {program: program_params}

    expect(response.status).to eq(422)
  end
end
