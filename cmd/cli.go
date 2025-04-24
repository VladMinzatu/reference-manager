package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/VladMinzatu/reference-manager/adapters"
	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "refman",
	Short: "Reference Manager CLI",
	Long:  "Command line interface for managing references organized by categories",
}

func main() {
	db, err := sql.Open("sqlite3", "db/references.db")
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	repo := adapters.NewSQLiteRepository(db)
	svc := service.NewReferenceService(repo)

	// Category commands
	var categoryCmd = &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
	}

	var addCategoryCmd = &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cat, err := svc.AddCategory(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Added category: %s (id: %d)\n", cat.Name, cat.Id)
			return nil
		},
	}

	var listCategoriesCmd = &cobra.Command{
		Use:   "list",
		Short: "List all categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, err := svc.GetAllCategories()
			if err != nil {
				return err
			}
			for _, cat := range categories {
				fmt.Printf("%d: %s\n", cat.Id, cat.Name)
			}
			return nil
		},
	}

	var deleteCategoryCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id format (must be integer): %v", err)
			}
			if err := svc.DeleteCategory(id); err != nil {
				return err
			}
			fmt.Printf("Deleted category with id: %d\n", id)
			return nil
		},
	}

	// Reference commands
	var referenceCmd = &cobra.Command{
		Use:   "reference",
		Short: "Manage references",
	}

	var listReferencesCmd = &cobra.Command{
		Use:   "list [categoryId]",
		Short: "List references in a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			refs, err := svc.GetReferences(categoryId)
			if err != nil {
				return err
			}
			for _, ref := range refs {
				ref.Render(&CLIReferenceRenderer{})
				fmt.Println()
			}
			return nil
		},
	}

	var addBookCmd = &cobra.Command{
		Use:   "add-book [categoryId] [title] [isbn] [description]",
		Short: "Add a book reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			book, err := svc.AddBookReference(categoryId, args[1], args[2], args[3])
			if err != nil {
				return err
			}
			fmt.Printf("Added book: %s (id: %d)\n", book.Title, book.Id)
			return nil
		},
	}

	var addLinkCmd = &cobra.Command{
		Use:   "add-link [categoryId] [title] [url] [description]",
		Short: "Add a link reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			link, err := svc.AddLinkReference(categoryId, args[1], args[2], args[3])
			if err != nil {
				return err
			}
			fmt.Printf("Added link: %s (id: %d)\n", link.Title, link.Id)
			return nil
		},
	}

	var addNoteCmd = &cobra.Command{
		Use:   "add-note [categoryId] [title] [text]",
		Short: "Add a note reference",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			note, err := svc.AddNoteReference(categoryId, args[1], args[2])
			if err != nil {
				return err
			}
			fmt.Printf("Added note: %s (id: %d)\n", note.Title, note.Id)
			return nil
		},
	}

	var deleteReferenceCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a reference",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid reference id: %v", err)
			}
			if err := svc.DeleteReference(id); err != nil {
				return err
			}
			fmt.Printf("Deleted reference with id: %d\n", id)
			return nil
		},
	}

	categoryCmd.AddCommand(addCategoryCmd, listCategoriesCmd, deleteCategoryCmd)
	referenceCmd.AddCommand(listReferencesCmd, addBookCmd, addLinkCmd, addNoteCmd, deleteReferenceCmd)
	rootCmd.AddCommand(categoryCmd, referenceCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type CLIReferenceRenderer struct{}

func (r *CLIReferenceRenderer) RenderBook(ref model.BookReference) {
	fmt.Printf("%d: [Book] %s\n", ref.Id, ref.Title)
	fmt.Printf("\t\t\tISBN: %s\n", ref.ISBN)
	fmt.Printf("\t\t\tDescription: %s\n", ref.Description)
}

func (r *CLIReferenceRenderer) RenderLink(ref model.LinkReference) {
	fmt.Printf("%d: [Link] %s\n", ref.Id, ref.Title)
	fmt.Printf("\t\t\tURL: %s\n", ref.URL)
	fmt.Printf("\t\t\tDescription: %s\n", ref.Description)
}

func (r *CLIReferenceRenderer) RenderNote(ref model.NoteReference) {
	fmt.Printf("%d: [Note] %s\n", ref.Id, ref.Title)
	fmt.Printf("\t\t\tText: %s\n", ref.Text)
}
