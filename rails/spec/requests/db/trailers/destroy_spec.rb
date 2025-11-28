# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/trailers/:id", type: :request do
  it "未ログインのとき、ログインページにリダイレクトすること" do
    trailer = FactoryBot.create(:trailer, :not_deleted)

    expect(Trailer.count).to eq(1)

    delete "/db/trailers/#{trailer.id}"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Trailer.count).to eq(1)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    trailer = FactoryBot.create(:trailer, :not_deleted)
    login_as(user, scope: :user)

    expect(Trailer.count).to eq(1)

    delete "/db/trailers/#{trailer.id}"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Trailer.count).to eq(1)
  end

  it "エディター権限を持つユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    trailer = FactoryBot.create(:trailer, :not_deleted)
    login_as(user, scope: :user)

    expect(Trailer.count).to eq(1)

    delete "/db/trailers/#{trailer.id}"
    trailer.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Trailer.count).to eq(1)
  end

  it "管理者権限を持つユーザーがログインしているとき、トレイラーを論理削除できること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    trailer = FactoryBot.create(:trailer, :not_deleted)
    login_as(user, scope: :user)

    expect(Trailer.count).to eq(1)

    delete "/db/trailers/#{trailer.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Trailer.count).to eq(0)
  end

  it "管理者権限を持つユーザーがログインしているとき、Refererがある場合はそのページにリダイレクトすること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    work = FactoryBot.create(:work)
    trailer = FactoryBot.create(:trailer, :not_deleted, work:)
    login_as(user, scope: :user)

    delete "/db/trailers/#{trailer.id}", headers: {"HTTP_REFERER" => "/db/works/#{work.id}/trailers"}

    expect(response).to redirect_to("/db/works/#{work.id}/trailers")
    expect(flash[:notice]).to eq("削除しました")
  end

  it "管理者権限を持つユーザーがログインしているとき、destroy_in_batchesメソッドが呼ばれること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    trailer = FactoryBot.create(:trailer, :not_deleted)
    login_as(user, scope: :user)

    # destroy_in_batchesメソッドが呼ばれることを確認
    trailer_relation = instance_double("ActiveRecord::Relation")
    allow(Trailer).to receive(:without_deleted).and_return(trailer_relation)
    allow(trailer_relation).to receive(:find).with(trailer.id.to_s).and_return(trailer)
    allow(trailer).to receive(:destroy_in_batches)

    delete "/db/trailers/#{trailer.id}"

    expect(trailer).to have_received(:destroy_in_batches)
    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
  end

  it "存在しないトレイラーIDを指定したとき、404エラーが返ること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    login_as(user, scope: :user)

    expect {
      delete "/db/trailers/non-existent-id"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みのトレイラーを削除しようとしたとき、404エラーが返ること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    trailer = FactoryBot.create(:trailer, :not_deleted)
    trailer.update!(deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      delete "/db/trailers/#{trailer.id}"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
