# typed: false
# frozen_string_literal: true

RSpec.describe "POST /api/internal/unlikes", type: :request do
  it "レビューのいいねを外した時、201ステータスを返すこと" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    work_record = create(:work_record, user: other_user)
    create(:like, user:, recipient: work_record)

    login_as(user, scope: :user)
    post "/api/internal/unlikes", params: {
      recipient_type: "WorkRecord",
      recipient_id: work_record.id
    }

    expect(response.status).to eq(201)
  end

  it "レビューのいいねを外した時、レコードが削除されること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    work_record = create(:work_record, user: other_user)
    create(:like, user:, recipient: work_record)

    login_as(user, scope: :user)
    expect {
      post "/api/internal/unlikes", params: {
        recipient_type: "WorkRecord",
        recipient_id: work_record.id
      }
    }.to change { user.likes.count }.from(1).to(0)
  end

  it "エピソードのいいねを外した時、201ステータスを返すこと" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    episode_record = create(:episode_record, user: other_user)
    create(:like, user:, recipient: episode_record)

    login_as(user, scope: :user)
    post "/api/internal/unlikes", params: {
      recipient_type: "EpisodeRecord",
      recipient_id: episode_record.id
    }

    expect(response.status).to eq(201)
  end

  it "エピソードのいいねを外した時、レコードが削除されること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    episode_record = create(:episode_record, user: other_user)
    create(:like, user:, recipient: episode_record)

    login_as(user, scope: :user)
    expect {
      post "/api/internal/unlikes", params: {
        recipient_type: "EpisodeRecord",
        recipient_id: episode_record.id
      }
    }.to change { user.likes.count }.from(1).to(0)
  end

  it "ログインしていない時、302ステータスを返すこと" do
    work_record = create(:work_record)

    post "/api/internal/unlikes", params: {
      recipient_type: "WorkRecord",
      recipient_id: work_record.id
    }

    expect(response.status).to eq(302)
  end

  it "存在しないレコードIDを指定した時、404ステータスを返すこと" do
    user = create(:registered_user)

    login_as(user, scope: :user)
    post "/api/internal/unlikes", params: {
    recipient_type: "WorkRecord",
    recipient_id: 999999

    expect(response.status).to eq(404)
  end

  it "いいねしていないレコードに対してリクエストした時、201ステータスを返すこと" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    work_record = create(:work_record, user: other_user)

    login_as(user, scope: :user)
    post "/api/internal/unlikes", params: {
      recipient_type: "WorkRecord",
      recipient_id: work_record.id
    }

    expect(response.status).to eq(201)
    expect(user.likes.count).to eq(0)
  end
end
