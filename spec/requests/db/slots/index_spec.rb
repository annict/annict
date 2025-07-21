# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/slots", type: :request do
  it "未ログインのとき、スロット一覧が表示されること" do
    slot = create(:slot)

    get "/db/works/#{slot.work_id}/slots"

    expect(response.status).to eq(200)
    expect(response.body).to include(slot.channel.name)
  end

  it "ログイン済みのとき、スロット一覧が表示されること" do
    user = create(:registered_user)
    slot = create(:slot)
    login_as(user, scope: :user)

    get "/db/works/#{slot.work_id}/slots"

    expect(response.status).to eq(200)
    expect(response.body).to include(slot.channel.name)
  end

  it "存在しない作品IDを指定したとき、404エラーが返ること" do
    expect do
      get "/db/works/non-existent-id/slots"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みの作品を指定したとき、404エラーが返ること" do
    work = create(:work, :deleted)

    expect do
      get "/db/works/#{work.id}/slots"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "program_idパラメータを指定したとき、該当するプログラムのスロットのみが表示されること" do
    work = create(:work)
    channel_group = ChannelGroup.create!(name: "地上波", sort_number: 1)
    channel1 = Channel.create!(channel_group:, name: "チャンネル1", sort_number: 1)
    channel2 = Channel.create!(channel_group:, name: "チャンネル2", sort_number: 2)
    program1 = create(:program, work:, channel: channel1)
    program2 = create(:program, work:, channel: channel2)
    slot1 = create(:slot, work:, program: program1, channel: channel1)
    slot2 = create(:slot, work:, program: program2, channel: channel2)

    get "/db/works/#{work.id}/slots", params: {program_id: program1.id}

    expect(response.status).to eq(200)
    # slot1のIDが表示されることを確認
    expect(response.body).to include(slot1.id.to_s)
    # slot2のIDが表示されないことを確認
    expect(response.body).not_to include(slot2.id.to_s)
  end

  it "削除済みのスロットは表示されないこと" do
    work = create(:work)
    slot = create(:slot, work:)
    deleted_slot = create(:slot, :deleted, work:)

    get "/db/works/#{work.id}/slots"

    expect(response.status).to eq(200)
    # 通常のスロットのエピソードタイトルが表示されること
    expect(response.body).to include(slot.episode.title)
    # 削除済みスロットのエピソードタイトルが表示されないこと
    expect(response.body).not_to include(deleted_slot.episode.title)
  end
end
