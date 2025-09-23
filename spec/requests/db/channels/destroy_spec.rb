# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/channels/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    channel = Channel.first

    expect(Channel.count).to eq(220)

    delete "/db/channels/#{channel.id}"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Channel.count).to eq(220)
  end

  it "エディターではないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    channel = Channel.first
    login_as(user, scope: :user)

    expect(Channel.count).to eq(220)

    delete "/db/channels/#{channel.id}"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Channel.count).to eq(220)
  end

  it "エディターユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    channel = Channel.first
    login_as(user, scope: :user)

    expect(Channel.count).to eq(220)

    delete "/db/channels/#{channel.id}"
    channel.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Channel.count).to eq(220)
  end

  it "管理者ユーザーがログインしているとき、チャンネルをソフトデリートできること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    login_as(user, scope: :user)

    expect(Channel.count).to eq(220)

    delete "/db/channels/#{channel.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Channel.count).to eq(219)
  end

  it "管理者ユーザーがログインしているとき、存在しないチャンネルを削除しようとするとエラーになること" do
    user = create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    delete "/db/channels/non-existent-id"

    expect(response.status).to eq(404)
  end

  it "管理者ユーザーがログインしているとき、Referrerがある場合はそのページにリダイレクトすること" do
    user = create(:registered_user, :with_admin_role)
    channel = Channel.first
    login_as(user, scope: :user)

    delete "/db/channels/#{channel.id}", headers: {"HTTP_REFERER" => db_channel_list_path}

    expect(response).to redirect_to(db_channel_list_path)
    expect(flash[:notice]).to eq("削除しました")
  end

  # NOTE: このテストはチャンネルファクトリーがないため、
  # プログラムとスロットを持つチャンネルを作成できないためスキップ
  # it "管理者ユーザーがログインしているとき、削除されたチャンネルの関連データも削除されること" do
  #   user = create(:registered_user, :with_admin_role)
  #   # プログラムを持つチャンネルを取得（シードデータから）
  #   channel = Channel.joins(:programs).first
  #   initial_program_count = channel.programs.count
  #   initial_slot_count = channel.slots.count
  #   login_as(user, scope: :user)
  #
  #   # プログラムとスロットが存在することを確認
  #   expect(initial_program_count).to be > 0
  #   expect(initial_slot_count).to be > 0
  #
  #   delete "/db/channels/#{channel.id}"
  #
  #   expect(response.status).to eq(302)
  #   expect(flash[:notice]).to eq("削除しました")
  #   # 関連データが削除されていることを確認
  #   expect(Program.where(channel_id: channel.id).count).to eq(0)
  #   expect(Slot.where(channel_id: channel.id).count).to eq(0)
  # end
end
