# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/slots/:id", type: :request do
  it "ログインしていないユーザーはアクセスできずログインページにリダイレクトされること" do
    channel = Channel.first
    slot = create(:slot)
    old_slot = slot.attributes
    slot_params = {
      channel_id: channel.id
    }

    patch "/db/slots/#{slot.id}", params: {slot: slot_params}
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(slot.channel_id).to eq(old_slot["channel_id"])
  end

  it "編集者権限のないユーザーはアクセスできないこと" do
    channel = Channel.first
    user = create(:registered_user)
    slot = create(:slot)
    old_slot = slot.attributes
    slot_params = {
      channel_id: channel.id
    }

    login_as(user, scope: :user)

    patch "/db/slots/#{slot.id}", params: {slot: slot_params}
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(slot.channel_id).to eq(old_slot["channel_id"])
  end

  it "編集者権限のあるユーザーがスロットを更新できること" do
    channel = Channel.first
    user = create(:registered_user, :with_editor_role)
    slot = create(:slot)
    old_slot = slot.attributes
    slot_params = {
      channel_id: channel.id
    }

    login_as(user, scope: :user)

    expect(slot.channel_id).to eq(old_slot["channel_id"])

    patch "/db/slots/#{slot.id}", params: {slot: slot_params}
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(slot.channel_id).to eq(channel.id)
  end

  it "編集者権限のあるユーザーが複数のパラメーターでスロットを更新できること" do
    channel = Channel.first
    user = create(:registered_user, :with_editor_role)
    slot = create(:slot)
    work = slot.work
    episode = create(:episode, work:)
    program = create(:program, work:)
    started_at = Time.zone.parse("2024-01-15 19:00:00")

    slot_params = {
      channel_id: channel.id,
      episode_id: episode.id,
      program_id: program.id,
      started_at:,
      number: 5,
      rebroadcast: true,
      irregular: true
    }

    login_as(user, scope: :user)

    patch "/db/slots/#{slot.id}", params: {slot: slot_params}
    slot.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")
    expect(slot.channel_id).to eq(channel.id)
    expect(slot.episode_id).to eq(episode.id)
    expect(slot.program_id).to eq(program.id)
    expect(slot.started_at.strftime("%Y-%m-%d %H:%M:%S")).to eq(started_at.strftime("%Y-%m-%d %H:%M:%S"))
    expect(slot.number).to eq(5)
    expect(slot.rebroadcast).to be(true)
    expect(slot.irregular).to be(true)
  end
end
